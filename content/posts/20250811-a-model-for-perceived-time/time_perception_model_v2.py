import numpy as np
import matplotlib.pyplot as plt
import os

def model_time_perception_final(max_age=81):
    """
    Models the perceived passage of time using monthly steps to approximate
    the continuous model.

    Args:
        max_age (int): The maximum age to model.

    Returns:
        tuple: A tuple containing the perceptual mid-point age and the path to the saved plot.
    """
    # Use monthly steps for a finer-grained summation, starting from age 1.
    months = np.arange(1, (max_age * 12) + 1)
    ages = months / 12.0

    # The perceived value of each month is dt/t, where dt = 1/12
    perceived_value_of_month = (1/12.0) / ages

    # We are modeling the time perception from age 1 onwards.
    # The integral model is from 1 to L, which is ln(L).
    # Our discrete sum approximates that.
    # We need to find the age M where the sum from 1 to M is half the total sum from 1 to L.
    
    # Let's find the cumulative sum starting from age 1 (month 12).
    age_1_index = 11 # Index for the 12th month
    
    cumulative_sum_from_age_1 = np.cumsum(perceived_value_of_month[age_1_index:])
    total_sum_from_age_1 = cumulative_sum_from_age_1[-1]
    half_sum_from_age_1 = total_sum_from_age_1 / 2.0

    mid_point_index = np.where(cumulative_sum_from_age_1 >= half_sum_from_age_1)[0][0]
    
    # Get the age in years at that midpoint index
    perceptual_mid_point_age = ages[age_1_index + mid_point_index]

    # Generate the visualization
    plot_ages = np.arange(1, max_age + 1)
    perceived_value_of_year = 1 / plot_ages

    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(10, 6))

    ax.plot(plot_ages, perceived_value_of_year, label='Perceived Value of a Year (1/age)')
    ax.axvline(perceptual_mid_point_age, color='r', linestyle='--', 
               label=f'Perceptual Mid-Point: Age {perceptual_mid_point_age:.1f}')

    ax.set_title('Model of Perceived Value of a Year vs. Age', fontsize=16)
    ax.set_xlabel('Age (Years)', fontsize=12)
    ax.set_ylabel('Perceived Value of a Single Year', fontsize=12)
    ax.legend()
    ax.grid(True)

    script_dir = os.path.dirname(os.path.abspath(__file__))
    plot_filename = os.path.join(script_dir, 'perceived_time_vs_age_v2.png')
    plt.savefig(plot_filename)
    plt.close()

    return perceptual_mid_point_age, plot_filename

if __name__ == '__main__':
    mid_point_age, plot_file = model_time_perception_final()
    print(f"The calculated perceptual mid-point of life is at age {mid_point_age:.1f}.")
    print(f"The visualization has been saved to '{plot_file}'.")
