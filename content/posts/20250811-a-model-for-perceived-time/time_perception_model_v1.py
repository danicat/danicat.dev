import numpy as np
import matplotlib.pyplot as plt
import os

def model_time_perception_v1(max_age=81):
    """
    Models the perceived passage of time using a simple discrete summation of years.
    This version is intentionally simple and has a known discrepancy with the continuous model.
    """
    ages = np.arange(1, max_age + 1)
    perceived_value_of_year = 1 / ages
    cumulative_perceived_time = np.cumsum(perceived_value_of_year)
    total_perceived_life_experience = cumulative_perceived_time[-1]
    half_perceived_life_experience = total_perceived_life_experience / 2
    mid_point_index = np.where(cumulative_perceived_time >= half_perceived_life_experience)[0][0]
    perceptual_mid_point_age = ages[mid_point_index]

    plt.style.use('seaborn-v0_8-whitegrid')
    fig, ax = plt.subplots(figsize=(10, 6))
    ax.plot(ages, perceived_value_of_year, label='Perceived Value of a Year (1/age)')
    ax.axvline(perceptual_mid_point_age, color='r', linestyle='--', label=f'Perceptual Mid-Point: Age {perceptual_mid_point_age}')
    ax.set_title('Model of Perceived Value of a Year vs. Age (V1: Annual Steps)', fontsize=16)
    ax.set_xlabel('Age (Years)', fontsize=12)
    ax.set_ylabel('Perceived Value of a Single Year', fontsize=12)
    ax.legend()
    ax.grid(True)

    script_dir = os.path.dirname(os.path.abspath(__file__))
    plot_filename = os.path.join(script_dir, 'perceived_time_vs_age_v1.png')
    plt.savefig(plot_filename)
    plt.close()

    return perceptual_mid_point_age, plot_filename

if __name__ == '__main__':
    mid_point_age, plot_file = model_time_perception_v1()
    print(f"V1 Model - Calculated perceptual mid-point: Age {mid_point_age}.")
    print(f"V1 plot saved to '{plot_file}'.")
